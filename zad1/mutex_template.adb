with Ada.Text_IO; use Ada.Text_IO;
with Ada.Numerics.Float_Random; use Ada.Numerics.Float_Random;
with Random_Seeds; use Random_Seeds;
with Ada.Real_Time; use Ada.Real_Time;
with Ada.Text_IO, Ada.Exceptions;
use Ada.Text_IO, Ada.Exceptions;

procedure  Mutex_Template is

-- Processes 

  Nr_Of_Processes : constant Integer :=15;

  Min_Steps : constant Integer := 10 ;
  Max_Steps : constant Integer := 30 ;

  Min_Delay : constant Duration := 0.01;
  Max_Delay : constant Duration := 0.05;

-- States of a Process 

  type Process_State is (
    Local_Section,
    Entry_Protocol_1,
    Entry_Protocol_2,
    Entry_Protocol_3,
    Entry_Protocol_4,
    Critical_Section,
    Exit_Protocol
    );

-- 2D Board display board

  Board_Width  : constant Integer := Nr_Of_Processes;
  Board_Height : constant Integer := Process_State'Pos( Process_State'Last ) + 1;

-- Timing

  Start_Time : Time := Clock;  -- global startnig time

-- Random seeds for the tasks' random number generators
 
  Seeds : Seed_Array_Type( 1 ..Nr_Of_Processes ) := Make_Seeds( Nr_Of_Processes );

  type Atomic_Boolean is new Boolean with Atomic, Volatile;
  type Atomic_Integer is new Integer with Atomic, Volatile;
  type Atomic_Integer_Array is array (0 .. Nr_Of_Processes-1) of Atomic_Integer;

  Flags : Atomic_Integer_Array; -- tickets for the ENTRY_PROTOCOL

   --  function Find_Max (Arr : Atomic_Integer_Array) return Atomic_Integer is
   --     Max : Atomic_Integer := Arr(Arr'First);
   --  begin
   --     for I in Arr'Range loop
   --        if Arr(I) > Max then
   --           Max := Arr(I); -- Atomic read of element
   --        end if;
   --     end loop;
   --     return Max;
   --  end Find_Max;

   --  function Trim_Image(I : Integer) return String is
   --  S : constant String := Integer'Image(I);
   --  begin
   --     return S(S'First + 1 .. S'Last);  -- skips leading space
   --  end;

-- Types, procedures and functions

  -- Postitions on the board
  type Position_Type is record	
    X: Integer range 0 .. Board_Width - 1; 
    Y: Integer range 0 .. Board_Height - 1; 
  end record;	   

  -- traces of Processes
  type Trace_Type is record 	      
    Time_Stamp:  Duration;	      
    Id : Integer;
    Position: Position_Type;      
    Symbol: Character;	      
  end record;	      

  type Trace_Array_type is  array(0 .. Max_Steps * 7) of Trace_Type;

  type Traces_Sequence_Type is record
    Last: Integer := -1;
    Trace_Array: Trace_Array_type ;
  end record; 

  procedure Print_Trace( Trace : Trace_Type ) is
    Symbol : String := ( ' ', Trace.Symbol );
  begin
    Put_Line(
        Duration'Image( Trace.Time_Stamp ) & " " &
        Integer'Image( Trace.Id ) & " " &
        Integer'Image( Trace.Position.X ) & " " &
        Integer'Image( Trace.Position.Y ) & " " &
        ( ' ', Trace.Symbol ) -- print as string to avoid: '
      );
  end Print_Trace;

  procedure Print_Traces( Traces : Traces_Sequence_Type ) is
  begin
    for I in 0 .. Traces.Last loop
      Print_Trace( Traces.Trace_Array( I ) );
    end loop;
  end Print_Traces;


  -- task Printer collects and prints reports of traces and the line with the parameters

  task Printer is
    entry Report( Traces : Traces_Sequence_Type );
  end Printer;
  
  task body Printer is 
 
  begin
      -- Collect and print the traces  
    for I in 1 .. Nr_Of_Processes loop -- range for TESTS !!!
        accept Report( Traces : Traces_Sequence_Type ) do
          -- Put_Line("I = " & I'Image );
          Print_Traces( Traces );
        end Report;
      end loop;

    -- Prit the line with the parameters needed for display script:

    Put(
      "-1 "&
      Integer'Image( Nr_Of_Processes ) &" "&
      Integer'Image( Board_Width ) &" "&
      Integer'Image( Board_Height ) &" "       
    );
    for I in Process_State'Range loop
      Put( I'Image &";" );
    end loop;
   
  end Printer;


  -- Processes
  type Process_Type is record
    Id: Integer;
    Symbol: Character;
    Position: Position_Type;
    ticket: Integer;
  end record;


  task type Process_Task_Type is	
    entry Init(Id: Integer; Seed: Integer; Symbol: Character);
    entry Start;
  end Process_Task_Type;	

  task body Process_Task_Type is
    G : Generator;
    Process : Process_Type;
    Time_Stamp : Duration;
    Nr_of_Steps: Integer;
    Traces: Traces_Sequence_Type;
    Did_I_Wait: Boolean;
    Found_Any: Boolean;

    procedure Store_Trace is
    begin  
      Traces.Last := Traces.Last + 1;
      Traces.Trace_Array( Traces.Last ) := ( 
          Time_Stamp => Time_Stamp,
          Id => Process.Id,
          Position => Process.Position,
          Symbol => Process.Symbol
        );
    end Store_Trace;

    procedure Change_State( State: Process_State ) is
    begin
      --  Put_Line ("Process " & Integer'Image(Process.Id) & " changed to " & Process_State'Image(State) );
      Time_Stamp := To_Duration ( Clock - Start_Time ); -- reads global clock
      Process.Position.Y := Process_State'Pos( State );
      Store_Trace;
    end;
    

  begin
    accept Init(Id: Integer; Seed: Integer; Symbol: Character) do
      Put_Line ("Initing task" & Integer'Image(Id));
      Reset(G, Seed); 
      Process.Id := Id;
      Process.Symbol := Symbol;
      Process.ticket := 0;
      -- Initial position 
      Process.Position := (
          X => Id,
          Y => Process_State'Pos( LOCAL_SECTION )
        );
      -- Number of steps to be made by the Process  
      Nr_of_Steps := Min_Steps + Integer( Float(Max_Steps - Min_Steps) * Random(G));
      -- Time_Stamp of initialization
      Time_Stamp := To_Duration ( Clock - Start_Time ); -- reads global clock
      Store_Trace; -- store starting position
      Did_I_Wait := False;
      Flags(Process.Id) := 0;
    end Init;
    
    -- wait for initialisations of the remaining tasks:
    accept Start do
      null;
    end Start;

--    for Step in 0 .. Nr_of_Steps loop
    for Step in 0 .. Nr_of_Steps - 1  loop  -- TEST !!!
      -- LOCAL_SECTION - start
      delay Min_Delay+(Max_Delay-Min_Delay)*Duration(Random(G));
      -- LOCAL_SECTION - end

      Flags(Process.Id) := 1;
      Change_State( Entry_Protocol_1 ); -- starting ENTRY_PROTOCOL
      Did_I_Wait := True;
      while Did_I_Wait = True loop
         Did_I_Wait := False;
         for I in Flags'Range loop
            while Flags(I) > 2 and Flags(I) < 5 loop
               Did_I_Wait := True;
               delay 0.0;
            end loop;
         end loop;
      end loop;

      Flags(Process.Id) := 3;
      Change_State ( Entry_Protocol_3);
      Found_Any := False;
      for I in Flags'Range loop
         if Flags(I) = 1 then
            Found_Any := True;
            --  Put_Line ("Exiting For");
            exit;
         end if;
      end loop;

      if Found_Any = True then
         Flags(Process.Id) := 2;
         Change_State ( Entry_Protocol_2 );
         -- await(any flag[1..N] = 4) 
         Found_Any := False;
         while Found_Any = False loop
            for I in Flags'Range loop
               if Flags(I) = 4 then
                  Found_Any := True;
                  --  Put_Line ("Exiting");
                  exit;
               end if;
            end loop;
         end loop;
      end if;

      Flags(Process.Id) := 4;
      Change_State (Entry_Protocol_4);
      -- await(all flag[1..self-1] ∈ {0, 1})
      --  Did_I_Wait := True;
      --  while Did_I_Wait = True loop
      --     Did_I_Wait := False;
      for I in 0 .. Process.Id - 1 loop
         while Flags(I) > 1 and Flags(I) < 5 loop
            --  Did_I_Wait := True;
            delay 0.0;
         end loop;
      end loop;
      --     end loop;
      --  end loop;

      
      Change_State( CRITICAL_SECTION ); -- starting CRITICAL_SECTION

      -- CRITICAL_SECTION - start
      delay Min_Delay+(Max_Delay-Min_Delay)*Duration(Random(G));
      -- CRITICAL_SECTION - end

      Change_State( EXIT_PROTOCOL ); -- starting EXIT_PROTOCOL
      if Process.Id < 14 then
         for I in Process.Id + 1 .. Flags'Last loop
            while Flags(I) > 1 and Flags(I) < 4 loop
               delay 0.0;
            end loop;
         end loop;
      end if;
      Flags(Process.Id) := 0;

      Change_State( LOCAL_SECTION ); -- starting LOCAL_SECTION      
    end loop;
    Put_Line ("Finished Steps");
    Flags(Process.Id) := 5; 
    -- Wait until all tasks finish
    for I in Flags'Range loop
      while Flags(I) < 5 loop
         delay 0.0;
      end loop;
    end loop;
    Put_Line ("Before sending report");
    Printer.Report( Traces );
    
    exception
   when E : others =>
      Put_Line("Task raised exception: " & Exception_Information(E));

  end Process_Task_Type;


-- local for main task

  Process_Tasks: array (0 .. Nr_Of_Processes-1) of Process_Task_Type; -- for tests
  Symbol : Character := 'A';

begin 

  -- init tasks
  for I in Process_Tasks'Range loop
    Process_Tasks(I).Init( I, Seeds(I+1), Symbol );   -- `Seeds(I+1)` is ugly :-(
    Symbol := Character'Succ( Symbol );
  end loop;

  -- start tarvelers tasks
  for I in Process_Tasks'Range loop
    Process_Tasks(I).Start;
  end loop;

end Mutex_Template;

